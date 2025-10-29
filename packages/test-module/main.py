import typing as t
from fastapi import FastAPI, Header

import requests

import krinon_models as s

app = FastAPI()


@app.get("/")
def home(krinon_headers: t.Annotated[s.KrinonHeaders, Header()]) -> s.AuthenticationInformation:
    public_key = requests.get("http://127.0.0.1:8000/.well-known/krinon-public-key").json()["key"]
    return s.AuthenticationInformation.from_jwt(krinon_headers.x_krinon_jwt, public_key)


