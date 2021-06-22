
from deta import Deta
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from datetime import datetime, timezone

deta = Deta() # project key is set automagically inside deta micros
app = FastAPI()
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
    allow_credentials=False,
)

@app.get("/")
async def root():
    return {"message": "hello world"}

class Payload(BaseModel):
    tz: str
    ua: str
    re: str
    ho: str
    pa: str

@app.post("/")
async def post(payload: Payload):
    record = {
            "timestamp": int(datetime.now(timezone.utc).timestamp()),
            "timezone": payload.tz,
            "useragent": payload.ua,
            "referrer": payload.re,
            "host": payload.ho,
            "path": payload.pa,
    }
    db = deta.Base("requests")
    db.put(record)
    return {"status": "ok"}
