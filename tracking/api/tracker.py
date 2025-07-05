from typing import List

from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from tracking.db.crud import get_trackers, get_tracker_by_id, update_tracker_resource
from tracking.dependencies import get_db
from tracking.models import TrackerModel, TrackerUpdateModel

tracker_router = APIRouter(
    prefix="/tracker", tags=["tracker"])


@tracker_router.get("/", response_model=List[TrackerModel])
def get_tracker(db: Session = Depends(get_db)):
    return get_trackers(db)


@tracker_router.post("/{instance_id}")
def update_tracker(instance_id: int, tracker_update_model: TrackerUpdateModel, db: Session = Depends(get_db)):
    tracker = get_tracker_by_id(db, instance_id)

    if tracker is None:
        raise HTTPException(status_code=404, detail="Couldn't create instance")

    tracker = update_tracker_resource(db, instance_id, tracker_update_model)

    return tracker
