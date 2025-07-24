from datetime import datetime
from typing import Union, List

from sqlalchemy import select
from sqlalchemy.orm import Session

from tracking.db.models import Tracker, Operation, Resource
from tracking.models import ChirpstackUpEventModel, ChirpstackPayloadBatteryMessage, ChirpstackPayloadLongitudeMessage, \
    ChirpstackPayloadLatitudeMessage, MCPTablueItem, TrackerUpdateModel


def get_tracker_by_id(db: Session, instance_id: int) -> Union[None, Tracker]:
    return db.execute(select(Tracker).filter_by(id=instance_id)).one_or_none()


def get_tracker_by_eui(db: Session, eui: str) -> Union[None, Tracker]:
    return db.execute(select(Tracker).filter_by(deviceEUI=eui)).one_or_none()


def get_trackers(db: Session) -> List[Tracker]:
    return db.execute(select(Tracker)).scalars().all()


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


def update_tracker_resource(db: Session, instance_id: int, model: TrackerUpdateModel):
    tracker = get_tracker_by_id(db, instance_id)

    if tracker is None:
        return None

    if model.resource is None:
        tracker.resourceId = None
    else:
        resource = get_resource_by_id(db, model.resource)
        if resource is None:
            return None

        # Ensure each assignment is unique
        trackers = db.execute(select(Tracker).filter_by(resourceId=model.resource)).all()

        if len(trackers) > 0:
            # Raise an error if the resourceId is already in use with another tracker
            return None

        # Update the value
        tracker.resourceId = model.resource

    db.commit()
    db.refresh(tracker)

    return tracker


def get_operation_by_uid(db: Session, uid: str) -> Union[None, Operation]:
    return db.execute(select(Operation).filter_by(uid=uid)).one_or_none()


def get_resource_by_id(db: Session, instance_id: int) -> Union[None, Resource]:
    return db.execute(select(Resource).filter_by(id=instance_id)).one_or_none()


def get_resource_by_uid(db: Session, uid: str) -> Union[None, Resource]:
    return db.execute(select(Resource).filter_by(uid=uid)).one_or_none()


def get_resources(db: Session) -> List[Resource]:
    return db.execute(select(Resource)).scalars().all()


def create_resource(db: Session, model: MCPTablueItem) -> Resource:

    resource = Resource(
        uid=model.resource.id,
        name=model.resource.name,
        type=model.resource.type,
        status=model.status
    )

    db.add(resource)
    db.commit()
    db.refresh(resource)
    return resource


def update_resource(db: Session, model: MCPTablueItem) -> Union[None, Resource]:
    resource = get_resource_by_uid(db, model.resource.id)
    if resource is None:
        return None

    resource.name = model.resource.name
    resource.type = model.resource.type
    resource.status = model.status

    db.commit()
    db.refresh(resource)
    return resource
