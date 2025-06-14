from fastapi import APIRouter
from tracking.api.gateway import gateway_router
from tracking.api.system import system_router

api_router = APIRouter(
    prefix="/api"
)

api_router.include_router(gateway_router)
api_router.include_router(system_router)
