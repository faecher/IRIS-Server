from datetime import datetime

from sqlalchemy import Integer, String, Float

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


class Operation(Base):
    # MCP class
    __tablename__ = 'operations'

    id = Column(Integer, primary_key=True)


class Team(Base):
    # MCP class
    __tablename__ = 'tableuItems'

    id = Column(Integer, primary_key=True)
