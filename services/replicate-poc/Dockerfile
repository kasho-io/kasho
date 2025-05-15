FROM python:3.13-slim

WORKDIR /app

# Install Poetry
RUN pip install poetry

# Copy poetry configuration files
COPY pyproject.toml poetry.lock* ./

# Configure poetry to not use a virtual environment
RUN poetry config virtualenvs.create false

# Install dependencies
RUN poetry install --no-root --no-interaction --no-ansi

# Copy application code
COPY . .

# Run the application
CMD ["python", "translicate_consumer.py"] 