from pydantic import BaseModel


class Capability(BaseModel):
    name: str
    module_id: int

class CapabilityCheckRequest(BaseModel):
    scope_id: int
    username: str
    capabilities: list[Capability]

class CapabilityCheckResponse(BaseModel):
    allowed_capabilities: list[Capability]
    denied_capabilities: list[Capability]