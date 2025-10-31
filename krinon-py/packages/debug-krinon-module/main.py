import typing as t
from fastapi import FastAPI, Header, Request
from fastapi.responses import HTMLResponse

import requests

import krinon_models as s

app = FastAPI()

@app.get("/module/path")
def module_path(krinon_headers: t.Annotated[s.KrinonHeaders, Header()]) -> str:
    public_key = requests.get("http://127.0.0.1/.well-known/krinon-public-key").text

    a = s.KrinonJWT.from_jwt(krinon_headers.x_krinon_jwt, public_key)

    return f"Yup, you got to module path as '{a.user_id}' in scope '{a.scope_id}'"

@app.get("/{full_path:path}")
def home(request: Request) -> HTMLResponse:
    print(request.headers)
    public_key = requests.get("http://127.0.0.1/.well-known/krinon-public-key").text

    krinon_jwt = request.headers.get("x-krinon-jwt")
    auth_info = s.KrinonJWT.from_jwt(krinon_jwt, public_key) if krinon_jwt else None
    return HTMLResponse(
        f"<p>The module sees that the request is for '{request.url}'."
         + (f"<br>Authenticated as '{auth_info.user_id}'" if auth_info and auth_info.user_id else "<br>Unauthenticated.")
         + (f"<br>In scope '{auth_info.scope_id}'</p>" if auth_info and auth_info.scope_id else "<br>Not in any scope.</p>"))

