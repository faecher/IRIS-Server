from fastapi import APIRouter
from tracking.api.gateway import gateway_router
from tracking.api.mcp import mcp_router
from tracking.api.resources import resource_router
from tracking.api.system import system_router
from tracking.api.tracker import tracker_router

api_router = APIRouter(
    prefix="/api"
)

api_router.include_router(gateway_router)
api_router.include_router(mcp_router)
api_router.include_router(resource_router)
api_router.include_router(system_router)
api_router.include_router(tracker_router)
