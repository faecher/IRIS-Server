import requests

from tracking import Settings
from tracking.db.models import Operation
from tracking.dependencies import get_db


async def get_mcp_data():
    settings = Settings()

    # Get the database
    db = next(get_db())

    if settings.mcp_enabled is True and settings.mcp_url is not None and settings.mcp_api_key is not None:
        print("Getting MCP data")

        # Check if there are any operations stored in the database
        # This should only be executed once on start
        operations = db.query(Operation).all()

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
            operation = db.query(Operation).filter(Operation.active == True).filter(Operation.selected == True).one_or_none()

            if operation is not None:
                request = requests.get(f"{settings.mcp_url}/api/tableau/resources",
                                       params={"operationId": operation.uid},
                                       headers={"Api-Key": settings.mcp_api_key, "accept": "*/*"},
                                       verify=False)
                if request.status_code < 400:
                    print(request.json())
                    # TODO: Save data to the database

    else:
        print("Skipping MPC query due to missing configuration")
