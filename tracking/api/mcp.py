import os

import requests
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from tracking import Settings
from tracking.db.crud import get_operation_by_uid
from tracking.db.models import Operation
from tracking.dependencies import get_db
from tracking.models import MCPConfig, MCPOperationConfig

mcp_router = APIRouter(
    prefix="/mcp", tags=["mcp"])


@mcp_router.get("/operations")
async def get_operations(db: Session = Depends(get_db)):
    operations = db.query(Operation).filter(Operation.active == True).all()

    return operations


@mcp_router.post("/operations/enable")
async def enable_operation(mcp_operation:MCPOperationConfig, db: Session = Depends(get_db)):

    operation = get_operation_by_uid(db, mcp_operation.uid)

    if operation is None:
        raise HTTPException(status_code=404, detail="Operation not found")
    else:
        operation.selected = True
        db.commit()
        db.refresh(operation)

    return {"status": 200}


@mcp_router.post("/operations/disable")
async def disable_operation(mcp_operation: MCPOperationConfig, db: Session = Depends(get_db)):
    operation = get_operation_by_uid(db, mcp_operation.uid)

    if operation is None:
        raise HTTPException(status_code=404, detail="Operation not found")
    else:
        operation.selected = False
        db.commit()
        db.refresh(operation)

    return {"status": 200}


@mcp_router.post("/start")
async def start_mcp(mcp_config: MCPConfig):

    # Sanitize the input
    if not mcp_config.url.startswith("https://"):
        mcp_config.url = f"https://{mcp_config.url}"

    if mcp_config.url.endswith("/"):
        mcp_config.url = mcp_config.url[:-1]

    # Try connection to the MCP server
    request = requests.get(f"{mcp_config.url}/api/version", verify=False, timeout=5)

    if request.status_code < 400:
        # Set the API details
        os.environ['MCP_URL'] = mcp_config.url
        os.environ['MCP_API_KEY'] = mcp_config.api_key
        os.environ['MCP_ENABLED'] = "true"
    else:
        raise HTTPException(status_code=400, detail="MCP server not running")

    return {"status": 200}


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

