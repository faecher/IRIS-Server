from fastapi import APIRouter

system_router = APIRouter(
    prefix="/system", tags=["system"])


@system_router.get("/status")
def get_system_status():
    return {}


@system_router.get("/version")
def get_system_version():
    return {"version": "beta"}
