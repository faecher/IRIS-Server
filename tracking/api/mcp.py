from fastapi import APIRouter

mcp_router = APIRouter(
    prefix="/mcp", tags=["mcp"])

@mcp_router.get("/config")
def get_config():
    return {}


@mcp_router.post("/config")
def set_config():
    return {}

