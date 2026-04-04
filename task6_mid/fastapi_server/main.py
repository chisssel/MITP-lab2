from fastapi import FastAPI, Request
import time
import uuid
import asyncio

app = FastAPI()


@app.get("/ping")
async def ping():
    return {"message": "pong"}


@app.get("/json")
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


@app.post("/echo")
async def echo(request: Request):
    body = await request.json()
    return body


@app.get("/slow")
async def slow():
    await asyncio.sleep(0.05)
    return {"status": "ok"}


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="127.0.0.1", port=8000, workers=1)
