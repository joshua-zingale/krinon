import typing as t
from pydantic import BaseModel
import jwt

class KrinonHeaders(BaseModel):
    x_krinon_jwt: str

class _JwtEncodableModel(BaseModel):

    @classmethod
    def from_jwt(cls, token: str, public_key: str | bytes) -> t.Self:
        return cls(
            **jwt.decode(token, public_key, algorithms="RS256"),
        )
    
    def to_jwt(self, private_key: bytes | str) -> str:
        return jwt.encode(self.model_dump(), private_key, "RS256")
    
class KrinonJWT(_JwtEncodableModel):
    user_id: t.Optional[str] = None
    """The ID of the authenticated user."""
    scope_ids: t.Optional[list[str]] = None
    """The ancestory of the scope in which a request was made.
    The first scope_id is the first parent and the last scope_id is the
    scope in which the request was made."""
    
class KrinonPublicKeyResponse(BaseModel):
    key: str