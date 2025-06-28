from datetime import datetime
from typing import Union

from pydantic import BaseModel, Field


class DeviceInfo(BaseModel):
    tenantId: str
    tenantName: str
    applicationId: str
    applicationName: str
    deviceProfileId: str
    deviceProfileName: str
    deviceName: str
    devEui: str
    # deviceClassEnabled seen in real data, but it does not show up in the docs
    # ignoring field tags


class ChirpstackPayloadMessage(BaseModel):
    type: str  # TODO: Replace by an enum or other better modeling technique
    timestamp: float
    measurementId: str


class ChirpstackPayloadBatteryMessage(BaseModel):
    type: str = Field(pattern="^Battery$")
    timestamp: float
    measurementId: str
    measurementValue: float


class ChirpstackPayloadLatitudeMessage(BaseModel):
    type: str = Field(pattern="^Latitude$")
    timestamp: float
    measurementId: str
    measurementValue: float


class ChirpstackPayloadLongitudeMessage(BaseModel):
    type: str = Field(pattern="^Longitude$")
    timestamp: float
    measurementId: str
    measurementValue: float


class ChirpstackPayloadObject(BaseModel):
    valid: bool
    payload: str
    err: float
    # WTF, why is the decoder designed that way?
    messages: list[list[Union[ChirpstackPayloadBatteryMessage, ChirpstackPayloadLatitudeMessage, ChirpstackPayloadLongitudeMessage, ChirpstackPayloadMessage]]]


class ChirpstackBaseEventModel(BaseModel):
    time: datetime
    deviceInfo: DeviceInfo


class ChirpstackUpEventModel(ChirpstackBaseEventModel):
    deduplicationId: str
    devAddr: str
    # ignoring a lot of other attributes
    object: ChirpstackPayloadObject


class TrackerModel(BaseModel):
    id: int
    deviceEUI: str
    name: str
    battery: float
    long: float
    lat: float
    lastUpdated: int

    class Config:
        from_attributes = True


class MCPConfig(BaseModel):
    url: str
    api_key: str


class MCPOperationConfig(BaseModel):
    uid: str
