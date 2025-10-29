import typing as t
from fastapi import FastAPI, Header
from pydantic import BaseModel

import requests

import krinon_models as s

app = FastAPI()

class DebugResponse(BaseModel):
    module_path: str
    authentication_information: s.AuthenticationInformation


@app.get("/module/path")
def module_path(krinon_headers: t.Annotated[s.KrinonHeaders, Header()]) -> str:
    public_key = requests.get("http://127.0.0.1:8070/.well-known/krinon-public-key").text

    a = s.AuthenticationInformation.from_jwt(krinon_headers.x_krinon_jwt, public_key)

    return f"Yup, you got to module path as '{a.user_id}' in scope '{a.scope_id}'"

@app.get("/{full_path:path}")
def home(full_path: str, krinon_headers: t.Annotated[s.KrinonHeaders, Header()]) -> DebugResponse:
    public_key = requests.get("http://127.0.0.1:8070/.well-known/krinon-public-key").text
    return DebugResponse(
        module_path=full_path,
        authentication_information=s.AuthenticationInformation.from_jwt(krinon_headers.x_krinon_jwt, public_key)
    )


