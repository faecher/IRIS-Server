# Addition of new Tracker Technologies

If you want to implement new tracker types and technologies, you mainly need to follow these three steps:

1. Add your data definition
   1. Add a table with your tracker identification data to the SQL scheme
   2. Add `/internal/models/[your Technology].go` with the corresponding data definition
2. Add your tracker into normal code handling:
   1. Create a function in `/internal/repository/trackerCreation.go` to create an instance of your tracker type in the db
   2. Adjust `GetAllTrackers()` in `/internal/repository/trackers.go` to also be able to handle your tracker technology
3. Implement your Tracker updates:
   1. create a submodule in /internal/ for your technology-specific code.
   2. Call this code either from a handler (if you get you data pushed to a webhook) or create a goroutine in `/cmd/iris-server/iris-server.go` that calls your polling-code
   

You can always look at Chirpstack / Tetra Implementations for reference.