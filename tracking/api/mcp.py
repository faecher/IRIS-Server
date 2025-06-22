from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session

from tracking import Settings
from tracking.db.models import Operation
from tracking.dependencies import get_db

mcp_router = APIRouter(
    prefix="/mcp", tags=["mcp"])


@mcp_router.get("/operations")
async def get_operations(db: Session = Depends(get_db)):
    operations = db.query(Operation).filter(Operation.active == True).all()

    return operations


@mcp_router.get("/config")
async def get_config(db: Session = Depends(get_db)):
    settings = Settings()

    operation = db.query(Operation).filter(Operation.active == True).filter(Operation.selected == True).one_or_none()

    result = {
        "enabled": settings.mcp_enabled,
        "api_key": settings.mcp_api_key,
        "url": settings.mcp_url,
        "operation_selected": operation is not None,
        "operation": operation.uid if operation is not None else ""
    }

    return result


@mcp_router.post("/config")
def set_config():
    return {}

