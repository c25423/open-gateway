# Use official Python 3.12 slim image based on Debian Bookworm
FROM python:3.12-slim-bookworm

# Set working directory
WORKDIR /app

# Copy requirements first to leverage Docker cache
COPY requirements.txt .

# Install dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code and configuration files
COPY src/ ./src/
COPY config.yaml ./

# Set default PORT environment variable
ENV HOST=0.0.0.0
ENV PORT=5000

# Set the command to run the gateway
CMD ["sh", "-c", "uvicorn src.server:app --host $HOST --port $PORT"]
