from datetime import datetime

from sqlalchemy.orm import Session

from tracking.db.models import Tracker
from tracking.models import ChirpstackUpEventModel, ChirpstackPayloadBatteryMessage, ChirpstackPayloadLongitudeMessage, \
    ChirpstackPayloadLatitudeMessage


def get_tracker_by_id(db: Session, instance_id: int):
    return db.query(Tracker).filter(Tracker.id == instance_id).first()


def get_tracker_by_eui(db: Session, eui: str):
    return db.query(Tracker).filter(Tracker.deviceEUI == eui).first()


def create_tracker(db: Session, model: ChirpstackUpEventModel):
    tracker = Tracker(
        deviceEUI=model.deviceInfo.devEui,
        name=model.deviceInfo.deviceName
    )

    # Check if the request contains messages
    if len(model.object.messages) > 0 and len(model.object.messages[0]) > 0:
        for item in model.object.messages[0]:
            # Check for a battery message
            if isinstance(item, ChirpstackPayloadBatteryMessage):
                tracker.battery = item.measurementValue
            # Check for a location message
            if isinstance(item, ChirpstackPayloadLatitudeMessage):
                tracker.lat = item.measurementValue
            if isinstance(item, ChirpstackPayloadLongitudeMessage):
                tracker.long = item.measurementValue

    db.add(tracker)
    db.commit()
    db.refresh(tracker)
    return tracker


def update_tracker(db: Session, model: ChirpstackUpEventModel):
    tracker = get_tracker_by_eui(db, model.deviceInfo.devEui)
    if tracker is None:
        return None

    has_update = False

    # Check if the request contains messages
    if len(model.object.messages) > 0 and len(model.object.messages[0]) > 0:
        for item in model.object.messages[0]:
            # Check for a battery message
            if isinstance(item, ChirpstackPayloadBatteryMessage):
                if int(item.timestamp / 1000) > tracker.lastUpdated:
                    tracker.battery = item.measurementValue
                    has_update = True
            # Check for a location message
            if isinstance(item, ChirpstackPayloadLatitudeMessage):
                if int(item.timestamp / 1000) > tracker.lastUpdated:
                    tracker.lat = item.measurementValue
                    has_update = True
            if isinstance(item, ChirpstackPayloadLongitudeMessage):
                if int(item.timestamp / 1000) > tracker.lastUpdated:
                    tracker.long = item.measurementValue
                    has_update = True

    # Update timestamp, if there was a data update
    if has_update:
        tracker.lastUpdated = int(datetime.now().timestamp())

    db.commit()
    db.refresh(tracker)
    return tracker
