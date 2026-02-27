package activity

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/baselyne/agent-sandbox-platform/control-plane/internal/models"
	pb "github.com/baselyne/agent-sandbox-platform/control-plane/pkg/gen/activity/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Handler implements the ActivityServiceServer gRPC interface.
type Handler struct {
	pb.UnimplementedActivityServiceServer
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RecordAction(ctx context.Context, req *pb.RecordActionRequest) (*pb.RecordActionResponse, error) {
	r := req.GetRecord()
	if r == nil {
		return nil, status.Error(codes.InvalidArgument, "record is required")
	}

	record := &models.ActionRecord{
		WorkspaceID:     r.GetWorkspaceId(),
		AgentID:         r.GetAgentId(),
		TaskID:          r.GetTaskId(),
		ToolName:        r.GetToolName(),
		Parameters:      structToJSON(r.GetParameters()),
		Result:          structToJSON(r.GetResult()),
		Outcome:         protoOutcomeToModel(r.GetOutcome()),
		GuardrailRuleID: r.GetGuardrailRuleId(),
		DenialReason:    r.GetDenialReason(),
	}

	if r.GetEvaluationLatencyUs() != 0 {
		v := r.GetEvaluationLatencyUs()
		record.EvaluationLatencyUs = &v
	}
	if r.GetExecutionLatencyUs() != 0 {
		v := r.GetExecutionLatencyUs()
		record.ExecutionLatencyUs = &v
	}

	id, err := h.svc.RecordAction(ctx, record)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.RecordActionResponse{RecordId: id}, nil
}

func (h *Handler) GetAction(ctx context.Context, req *pb.GetActionRequest) (*pb.GetActionResponse, error) {
	record, err := h.svc.GetAction(ctx, req.GetRecordId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	pbRecord, err := recordToProto(record)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "convert record: %v", err)
	}
	return &pb.GetActionResponse{Record: pbRecord}, nil
}

func (h *Handler) QueryActions(ctx context.Context, req *pb.QueryActionsRequest) (*pb.QueryActionsResponse, error) {
	var startTime, endTime *timestamppb.Timestamp
	startTime = req.GetStartTime()
	endTime = req.GetEndTime()

	filter := QueryFilter{
		WorkspaceID: req.GetWorkspaceId(),
		AgentID:     req.GetAgentId(),
		TaskID:      req.GetTaskId(),
		ToolName:    req.GetToolName(),
		Outcome:     protoOutcomeToModel(req.GetOutcome()),
		AfterID:     req.GetPageToken(),
		Limit:       int(req.GetPageSize()),
	}
	if startTime != nil {
		t := startTime.AsTime()
		filter.StartTime = &t
	}
	if endTime != nil {
		t := endTime.AsTime()
		filter.EndTime = &t
	}

	records, nextToken, err := h.svc.QueryActions(ctx, filter)
	if err != nil {
		return nil, toGRPCError(err)
	}

	pbRecords := make([]*pb.ActionRecord, len(records))
	for i := range records {
		pbRec, err := recordToProto(&records[i])
		if err != nil {
			return nil, status.Errorf(codes.Internal, "convert record: %v", err)
		}
		pbRecords[i] = pbRec
	}
	return &pb.QueryActionsResponse{
		Records:       pbRecords,
		NextPageToken: nextToken,
		TotalCount:    int32(len(pbRecords)),
	}, nil
}

func (h *Handler) StreamActions(req *pb.StreamActionsRequest, stream grpc.ServerStreamingServer[pb.ActionRecord]) error {
	workspaceFilter := req.GetWorkspaceId()
	agentFilter := req.GetAgentId()

	subID, ch := h.svc.Broker().Subscribe()
	defer h.svc.Broker().Unsubscribe(subID)

	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case record, ok := <-ch:
			if !ok {
				return nil
			}
			// Apply filters.
			if workspaceFilter != "" && record.WorkspaceID != workspaceFilter {
				continue
			}
			if agentFilter != "" && record.AgentID != agentFilter {
				continue
			}
			pbRecord, err := recordToProto(record)
			if err != nil {
				return status.Errorf(codes.Internal, "convert record: %v", err)
			}
			if err := stream.Send(pbRecord); err != nil {
				return err
			}
		}
	}
}

// --- converters ---

func recordToProto(r *models.ActionRecord) (*pb.ActionRecord, error) {
	params, err := jsonToStruct(r.Parameters)
	if err != nil {
		return nil, err
	}
	result, err := jsonToStruct(r.Result)
	if err != nil {
		return nil, err
	}

	rec := &pb.ActionRecord{
		RecordId:        r.ID,
		WorkspaceId:     r.WorkspaceID,
		AgentId:         r.AgentID,
		TaskId:          r.TaskID,
		ToolName:        r.ToolName,
		Parameters:      params,
		Result:          result,
		Outcome:         modelOutcomeToProto(r.Outcome),
		GuardrailRuleId: r.GuardrailRuleID,
		DenialReason:    r.DenialReason,
		RecordedAt:      timestamppb.New(r.RecordedAt),
	}
	if r.EvaluationLatencyUs != nil {
		rec.EvaluationLatencyUs = *r.EvaluationLatencyUs
	}
	if r.ExecutionLatencyUs != nil {
		rec.ExecutionLatencyUs = *r.ExecutionLatencyUs
	}
	return rec, nil
}

func protoOutcomeToModel(o pb.ActionOutcome) models.ActionOutcome {
	switch o {
	case pb.ActionOutcome_ACTION_OUTCOME_ALLOWED:
		return models.ActionOutcomeAllowed
	case pb.ActionOutcome_ACTION_OUTCOME_DENIED:
		return models.ActionOutcomeDenied
	case pb.ActionOutcome_ACTION_OUTCOME_ESCALATED:
		return models.ActionOutcomeEscalated
	case pb.ActionOutcome_ACTION_OUTCOME_ERROR:
		return models.ActionOutcomeError
	default:
		return ""
	}
}

func modelOutcomeToProto(o models.ActionOutcome) pb.ActionOutcome {
	switch o {
	case models.ActionOutcomeAllowed:
		return pb.ActionOutcome_ACTION_OUTCOME_ALLOWED
	case models.ActionOutcomeDenied:
		return pb.ActionOutcome_ACTION_OUTCOME_DENIED
	case models.ActionOutcomeEscalated:
		return pb.ActionOutcome_ACTION_OUTCOME_ESCALATED
	case models.ActionOutcomeError:
		return pb.ActionOutcome_ACTION_OUTCOME_ERROR
	default:
		return pb.ActionOutcome_ACTION_OUTCOME_UNSPECIFIED
	}
}

// structToJSON converts a proto Struct to json.RawMessage.
func structToJSON(s *structpb.Struct) json.RawMessage {
	if s == nil {
		return nil
	}
	b, err := s.MarshalJSON()
	if err != nil {
		return nil
	}
	return json.RawMessage(b)
}

// jsonToStruct converts json.RawMessage to a proto Struct.
func jsonToStruct(data json.RawMessage) (*structpb.Struct, error) {
	if len(data) == 0 {
		return nil, nil
	}
	s := &structpb.Struct{}
	if err := s.UnmarshalJSON(data); err != nil {
		return nil, err
	}
	return s, nil
}

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, ErrRecordNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, ErrInvalidInput):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
