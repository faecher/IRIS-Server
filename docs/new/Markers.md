First we need a place-ID. All available places can be found in `/api/places`.
the place ID can also be gotten from the `/api/operations`.
--> Frage: kann es mehr als eine aktive operation geben oder kann man da grade annehmen, dass die erste aktive Operation auch die einzige ist?

Dann brauchen wir einen Siteplan aus diesem Place.
Siehe `/api/siteplan/template?placeId=<placeID>`.
ein place kann viele siteplans haben, da brauchen wir auf jeden Fall einen Selector im UI.

mit der siteplan-id kann man marker erstellen, aber achtung: entityType muss TEMPLATE und nicht SNAPSHOT sein!
beim erstellen muss man zwar das id feld angeben für den marker,  bekommt dann aber eine komplett andere ID zurück für den kreierten Marker