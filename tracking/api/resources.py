from fastapi import APIRouter

resource_router = APIRouter(
    prefix="/resources", tags=["resources"])