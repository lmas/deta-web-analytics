
import ulid
from deta import Deta
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel

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
    ts: int
    tz: str
    ua: str
    ref: str
    url: str

@app.post("/post")
async def post(payload: Payload):
    record = {
            "key": ulid.from_timestamp(payload.ts).str,
            "timezone": payload.tz,
            "useragent": payload.ua,
            "referrer": payload.ref,
            "url": payload.url,
    }
    db = deta.Base("requests")
    db.put(record)
    return {"status": "ok"}
