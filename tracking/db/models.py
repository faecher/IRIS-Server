from datetime import datetime
from typing import Optional

from sqlalchemy import Integer, String, Float, Boolean, ForeignKey
from sqlalchemy.orm import relationship, mapped_column, Mapped

from tracking.db import Base, Column


class Tracker(Base):
    __tablename__ = 'trackers'

    id = Column(Integer, primary_key=True)
    deviceEUI = Column(String, unique=True)
    name = Column(String)
    battery = Column(Float, default=0.0)
    long = Column(Float, default=0.0)
    lat = Column(Float, default=0.0)
    # Use unix timestamp for last updated
    lastUpdated = Column(Integer, default=int(datetime.now().timestamp()))

    # Database relationship
    resourceId: Mapped[Optional[int]] = mapped_column(ForeignKey("resources.id"))
    resource: Mapped[Optional["Resource"]] = relationship(back_populates="tracker")


class Operation(Base):
    # MCP class
    __tablename__ = 'operations'

    id = Column(Integer, primary_key=True)

    uid = Column(String)
    title = Column(String)
    active = Column(Boolean)
    archived = Column(Boolean)
    # ignore all other attributes
    selected = Column(Boolean, default=False)


class Resource(Base):
    # MCP class
    __tablename__ = 'resources'

    id = Column(Integer, primary_key=True)

    # UUID of the resource
    uid = Column(String)

    name = Column(String)
    type = Column(String)
    status = Column(Integer)

    # Database relationship
    tracker: Mapped["Tracker"] = relationship(back_populates="resource")
