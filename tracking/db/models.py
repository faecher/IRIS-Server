from sqlalchemy import Integer, String, Float, DateTime
from sqlalchemy import func

from tracking.db import Base, Column


class Tracker(Base):
    __tablename__ = 'trackers'

    id = Column(Integer, primary_key=True)
    deviceEUI = Column(String, unique=True)
    name = Column(String)
    battery = Column(Float, default=0.0)
    long = Column(Float, default=0.0)
    lat = Column(Float, default=0.0)
    lastUpdated = Column(DateTime, server_default=func.current_timestamp())
