
from deta import Deta, App
from fastapi import FastAPI, Response
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from datetime import datetime, timezone, timedelta
from requests import get as httpGet

api = FastAPI()
api.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
    allow_credentials=False,
)
app = App(api)
deta = Deta() # project key is set automagically inside deta micros

source = "https://github.com/lmas/deta-web-analytics"
userAgent = "Mozilla/5.0 (compatible; PingBot/0.1; +" + source + ")"
pingTimeout = 10
targetHost = "www.larus.se"

def timestamp():
    return int(datetime.now(timezone.utc).timestamp())

####################################################################################################

@app.get("/")
async def index():
    # Be a better citizen and link to more info
    body = """Hello world!
For more info, see:
"""
    body = body + source
    return Response(content=body, media_type="text/plain")

@app.get("/robots.txt")
async def robots():
    # Try and stop bots/crawlers from bothering us
    body = """User-agent: *
Disallow: /
"""
    return Response(content=body, media_type="text/plain")

class Payload(BaseModel):
    tz: str
    ua: str
    re: str
    ho: str
    pa: str

@app.post("/ping")
async def post_ping(payload: Payload):
    # Recieves a ping analytics payload from some JS client/browser
    if payload.ho != targetHost:
        return {"status": "invalid host"}

    record = {
            "timestamp": timestamp(),
            "timezone": payload.tz,
            "useragent": payload.ua,
            "referrer": payload.re,
            "host": payload.ho,
            "path": payload.pa,
    }
    db = deta.Base("requests")
    db.put(record)
    return {"status": "post ok"}

@app.lib.cron()
def cron_ping(event):
    # Periodically ping the host and log it's response status/time
    resp = httpGet(
            "https://"+targetHost,
            headers={"User-Agent": userAgent},
            timeout=pingTimeout,
            allow_redirects=True,
            verify=True,
    )
    record = {
            "timestamp": timestamp(),
            "host": targetHost,
            "status": resp.status_code,
            "elapsed": resp.elapsed / timedelta(milliseconds=1),
    }
    db = deta.Base("responses")
    db.put(record)
    return {"status": "ping ok"}
