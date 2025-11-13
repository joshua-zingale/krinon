from fastapi import FastAPI, Request
from fastapi.responses import HTMLResponse

import requests

import krinon_models as s

app = FastAPI()

@app.get("/{full_path:path}")
def home(request: Request) -> HTMLResponse:
    print(request.headers)
    public_key = requests.get(f"http://{request.headers.get("x-forwarded-host", "127.0.0.1")}/.well-known/krinon-public-key").text

    krinon_jwt = request.headers.get("x-krinon-jwt")
    auth_info = s.KrinonJWT.from_jwt(krinon_jwt, public_key) if krinon_jwt else None
    return HTMLResponse(
        f"<p>The module sees that the request is for '{request.url}'."
         + (f"<br>Authenticated as '{auth_info.user_id}'" if auth_info and auth_info.user_id else "<br>Unauthenticated.")
         + (f"<br>In scope '{auth_info.scope_ids}'</p>" if auth_info and auth_info.scope_ids else "<br>Not in any scope.</p>"))
