FROM python:3.12-slim

WORKDIR /app

# Install the SDK.
COPY sdk/python /tmp/sdk
RUN pip install --no-cache-dir /tmp/sdk && rm -rf /tmp/sdk

COPY tests/fixtures/test-agent/agent.py .

ENTRYPOINT ["python", "agent.py"]
