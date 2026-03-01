package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// printTable writes rows as an aligned table to w.
func printTable(w io.Writer, headers []string, rows [][]string) {
	tw := tabwriter.NewWriter(w, 0, 8, 2, ' ', 0)
	fmt.Fprintln(tw, strings.Join(headers, "\t"))
	for _, row := range rows {
		fmt.Fprintln(tw, strings.Join(row, "\t"))
	}
	tw.Flush()
}

// printJSON marshals v as indented JSON to w.
func printJSON(w io.Writer, v any) error {
	// If it's a proto message, use protojson for canonical output.
	if msg, ok := v.(proto.Message); ok {
		opts := protojson.MarshalOptions{
			Multiline: true,
			Indent:    "  ",
		}
		b, err := opts.Marshal(msg)
		if err != nil {
			return err
		}
		_, err = w.Write(b)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(w)
		return err
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// outputResult dispatches to table or JSON based on the --output flag.
// For table output, pass headers and rows. For JSON, pass the proto message.
func outputResult(cmd *cobra.Command, headers []string, rows [][]string, jsonObj any) error {
	if getOutputFormat(cmd) == "json" {
		return printJSON(os.Stdout, jsonObj)
	}
	printTable(os.Stdout, headers, rows)
	return nil
}

// resolveID gets an ID from the first positional arg or the named flag.
func resolveID(cmd *cobra.Command, args []string, flagName string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}
	id, _ := cmd.Flags().GetString(flagName)
	if id == "" {
		return "", fmt.Errorf("provide ID as argument or --%s flag", flagName)
	}
	return id, nil
}

// formatTimestamp formats a proto timestamp for table display.
func formatTimestamp(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return "-"
	}
	return ts.AsTime().Format(time.RFC3339)
}
