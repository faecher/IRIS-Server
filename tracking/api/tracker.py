from typing import List

from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session

from tracking.db.crud import get_trackers
from tracking.dependencies import get_db
from tracking.models import TrackerModel

tracker_router = APIRouter(
    prefix="/tracker", tags=["tracker"])


@tracker_router.get("/", response_model=List[TrackerModel])
def get_tracker(db: Session = Depends(get_db)):
    return get_trackers(db)
