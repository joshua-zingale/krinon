from fastapi import FastAPI, Request
from fastapi.responses import HTMLResponse

import requests

import krinon_models as s

app = FastAPI()

@app.get("/{full_path:path}")
def home(request: Request) -> HTMLResponse:
    public_key = requests.get(f"http://{request.headers.get("x-forwarded-host", "127.0.0.1")}/.well-known/krinon-public-key").text

    krinon_jwt = request.headers.get("x-krinon-jwt")
    auth_info = None
    if krinon_jwt:
        try:
            auth_info = s.KrinonJWT.from_jwt(krinon_jwt, public_key, audiance={
                "host": str(request.url.hostname),
                "port": int(request.url.port or 80),
                "path": request.url.path})
        except Exception as e:
            return HTMLResponse(
                f"<p>Problem when gathering auth_info: {e}</p><br><p>headers: {request.headers}</p>"
            )
    return HTMLResponse(
        f"<p>The module sees that the request is for '{request.url}'."
         + (f"<br>Authenticated as '{auth_info.user_id}'" if auth_info and auth_info.user_id else "<br>Unauthenticated.")
         + (f"<br>In scope '{auth_info.scope_ids}'</p>" if auth_info and auth_info.scope_ids else "<br>Not in any scope.</p>")
         + (f"<br><br>auth_info: {auth_info}"))
