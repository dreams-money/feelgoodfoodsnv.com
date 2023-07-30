![Feel Good Foods Logo](https://user-images.githubusercontent.com/5210627/256919046-99a366da-361c-4c90-b316-1a6c9a0e438b.png)

Feel Good Foods
===============


A food ordering and delivery system.


Configuration
-------------
The system requires that delivery slots and a menu be defined before it can start processing orders.  As such, you'll
want to run `go run main.go` once on project initialization to initialize the `/data` folder.

With main.go running, navigate to http://\<project-url>/admin.  The login will be:

    admin@feelgoodfoodsnv.com
    asdfasdf

This is where you can set up the delivery slots and menu.  Once it's set up, restart main.go and the order manager will
get started.
