FROM python:3.11-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir --upgrade -r requirements.txt
COPY . .

ENV PORT=8000
EXPOSE ${PORT}
CMD ["fastapi", "run", "/app/tracking/__init__.py", "--port", "${PORT}"]