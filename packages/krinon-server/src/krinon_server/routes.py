from fastapi import APIRouter, Response, Request
from httpx import AsyncClient
from krinon_models import schemata as s

with open("../../../../private.pem") as f:
    PRIVATE_KEY = f.read()

with open("../../../../public.pem") as f:
    PUBLIC_KEY = f.read()

router = APIRouter()

client = AsyncClient()

@router.get("/.well-known/krinon-public-key")
async def get_krinon_public_key() -> s.KrinonPublicKeyResponse:
    return s.KrinonPublicKeyResponse(key=PUBLIC_KEY)

USER_ID = "tom@example.com"
SCOPE_ID = "example-scope"
MODULE_URL = "http://127.0.0.1:5555"
@router.route("/{full_path:path}", methods=["GET","POST","PUT","DELETE","PATCH"])
async def gateway_proxy(request: Request) -> Response:
    """Adds user and scope information to a request, extracts a module path, and
    resends to request to the module. The scope is the part of the url before "/m/"
    the module is the first name after it, and what follows is the path sent to the
    module. The response should be received from the module and then sent back
    to the original sender with its source header modified to the proxy path."""

    request_headers = request.headers.mutablecopy()
    request_headers.update({
        "X-Krinon-Jwt": s.AuthenticationInformation(
            user_id=USER_ID,
            scope_id=SCOPE_ID,
        ).to_jwt(PRIVATE_KEY)
    })

    response = await client.request(request.method, MODULE_URL, headers=request_headers)


    return Response(
        content=response.content,
        headers=response.headers)