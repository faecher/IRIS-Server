from sqlalchemy import select
import requests

from tracking import Settings
from tracking.db.crud import get_resource_by_uid, create_resource, update_resource
from tracking.db.models import Operation
from tracking.dependencies import get_db
from tracking.models import MCPTablueItem


async def get_mcp_data():
    settings = Settings()

    # Get the database
    db = await anext(get_db())

    if settings.mcp_enabled is True and settings.mcp_url is not None and settings.mcp_api_key is not None:
        print("Getting MCP data")

        # Check if there are any operations stored in the database
        # This should only be executed once on start
        operations = db.execute(select(Operation)).all()

        if len(operations) == 0:
            request = requests.get(f"{settings.mcp_url}/api/operations", headers={"Api-Key": settings.mcp_api_key, "accept": "*/*"},
                         verify=False)

            if request.status_code < 400:
                operations_json = request.json()
                # Save results to the database

                for item in operations_json:
                    operation = Operation(
                        uid=item['id'],
                        title=item['title'],
                        active=item['active'],
                        archived=item['archived']
                    )
                    db.add(operation)
                    db.commit()
        else:
            # Get the selected operation
            operation = db.execute(select(Operation).filter_by(active=True).filter_by(selected=True)).one_or_none()

            if operation is not None:
                request = requests.get(f"{settings.mcp_url}/api/tableau/resources",
                                       params={"operationId": operation.uid},
                                       headers={"Api-Key": settings.mcp_api_key, "accept": "*/*"},
                                       verify=False)
                if request.status_code < 400:
                    for item in request.json():
                        # Convert item to pydantic model
                        resource = MCPTablueItem.model_validate(item)

                        # Check if there is already such a resource
                        db_resource = get_resource_by_uid(db, resource.resource.id)

                        if db_resource is None:
                            # Create a new resource
                            create_resource(db, resource)
                        else:
                            update_resource(db, resource)

    else:
        print("Skipping MPC query due to missing configuration")
