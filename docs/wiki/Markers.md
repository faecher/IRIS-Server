# Exkurs: MCP Marker

## Über Marker:
Marker können mittels POST erstellt und mittels PUT bearbeitet werden

Bei der Bearbeitung ist nur die Marker-ID zur identifikation relevant, es müssen jedoch (scheinbar) immer alle Felder übermittelt werden.

Jeder marker ist nur auf seinem bestehenden siteplan existent.
Gibt man eine andere Siteplan-ID an, wird diese ignoriert und es kommt die alte zurück, das Feld muss aber trotzdem mitgeschickt werden -.-



## Um an alle Marker (Auf einem Siteplan) zu kommen
First we need a place-ID. All available places can be found in `/api/places`.
The place ID can also be gotten from the `/api/operations`, if you know which operation is at your desired place.

Dann brauchen wir einen Siteplan aus diesem Place.
Siehe `/api/siteplan/template?placeId=<placeID>`.
ein place kann viele siteplans haben -> da brauchen wir auf jeden Fall einen Selector im UI.

mit der siteplan-id kann man marker erstellen, aber achtung: entityType muss TEMPLATE und nicht SNAPSHOT sein!
beim erstellen muss man zwar das id feld angeben für den marker, bekommt dann aber eine komplett andere ID zurück für den kreierten Marker



## Bedeutet für IRIS
IRIS hat ein Verzeichnis aller von Iris angelegten Marker, mit zugehöriger Siteplan, Marker und (von IRIS zugeteilter) Ressourcen-ID.
Dementsprechend kann zu jedem Zeitpunkt für eine bestimmte Ressource der Marker bestimmt werden, der bewegt werden soll.



