
from fastapi import FastAPI, Body

app = FastAPI()

@app.get("/")
async def root():
    print("test")
    return {"message": "hello world"}

@app.post("/post")
async def post(ts: int, tz: str, ua: str, ref: str, url: str = Body(...)):
    print(ts, tz, ua, ref, url)
    return {"status": "ok"}
