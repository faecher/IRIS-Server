from typing import List

from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session

from tracking.db.crud import get_resources
from tracking.dependencies import get_db
from tracking.models import Resource

resource_router = APIRouter(
    prefix="/resources", tags=["resources"])


@resource_router.get("/", response_model=List[Resource])
def get_resource(db: Session = Depends(get_db)):
    return get_resources(db)
