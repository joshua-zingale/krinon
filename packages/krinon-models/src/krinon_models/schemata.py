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
    
class AuthenticationInformation(_JwtEncodableModel):
    user_id: str
    scope_id: str

    
class KrinonPublicKeyResponse(BaseModel):
    key: str