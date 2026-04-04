from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse
from pydantic import BaseModel, Field
import time
import uuid
from typing import List, Optional

app = FastAPI(
    title="FastAPI Server",
    description="REST API with Swagger documentation",
    version="1.0.0",
    docs_url="/docs",
    redoc_url="/redoc",
    openapi_url="/openapi.json"
)


class DataNested(BaseModel):
    key: str = Field(description="Nested key value")


class DataItems(BaseModel):
    items: List[str] = Field(description="List of items")
    count: int = Field(description="Number of items")
    nested: DataNested


class JsonResponse(BaseModel):
    status: str = Field(description="Response status")
    timestamp: int = Field(description="Unix timestamp in milliseconds")
    id: str = Field(description="Unique identifier (UUID)")
    data: DataItems


class EchoRequest(BaseModel):
    key: Optional[str] = None
    value: Optional[str] = None


class SlowResponse(BaseModel):
    status: str = Field(description="Response status")


class HealthResponse(BaseModel):
    status: str = Field(description="Health status")


@app.get(
    "/ping",
    response_model=dict,
    summary="Ping endpoint",
    description="Returns a simple pong message for health checks"
)
async def ping():
    return {"message": "pong"}


@app.get(
    "/json",
    response_model=JsonResponse,
    summary="JSON endpoint",
    description="Returns a complex JSON response with UUID and nested data",
    responses={
        200: {"description": "Successful response"}
    }
)
async def json_endpoint():
    return {
        "status": "ok",
        "timestamp": int(time.time() * 1000),
        "id": str(uuid.uuid4()),
        "data": {
            "items": ["item1", "item2", "item3"],
            "count": 3,
            "nested": {"key": "value"},
        },
    }


@app.post(
    "/echo",
    response_model=dict,
    summary="Echo endpoint",
    description="Returns the same JSON body that was sent in the request",
    responses={
        200: {"description": "Echoed body"},
        400: {"description": "Invalid JSON body"}
    }
)
async def echo(request: Request):
    body = await request.json()
    return body


@app.get(
    "/slow",
    response_model=SlowResponse,
    summary="Slow endpoint",
    description="Returns a response after a 50ms delay (simulates slow processing)"
)
async def slow():
    import asyncio
    await asyncio.sleep(0.05)
    return {"status": "ok"}


@app.get(
    "/health",
    response_model=HealthResponse,
    summary="Health check",
    description="Returns the health status of the server"
)
async def health():
    return {"status": "healthy"}


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="127.0.0.1", port=8000, workers=1)
