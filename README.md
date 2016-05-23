Programming Question
====================

We have the following scenario: Our traffic robots travel around London to
report on traffic conditions. Every time a robot passes close to a tube
station, it assesses the traffic condition in the area, and reports it.

Task:
-----

You should write a program that will have the following actors: one dispatcher
and two robots. Each of the robots should “move” on separate threads. The
dispatcher will be responsible for passing the points to the robot, and for
terminating the simulation. Upon receiving the new points, the robot moves,
checks if there is a station near by, and if so, reports on the traffic
condition. The information should be posted in such a way that is accessible to
any other component of the system.

Notes
-----

 + The robots cannot store lots of information, so they have to consume 10
   points at a time.
 + The simulation should end at 8:10 am, when the robots will receive a
   “SHUTDOWN” instruction.
 + There are two robots, identified by their id: 6403 and 5937. You will
   find a corresponding file with points (lat/lon) along their routes.
   The layout of the file is: driver id,latitude,longitude,time
 + There is also a file containg lat/lon for London's tube stations. The layout
   of the file is: description,lat,lon
 + The traffic information should have the following format:
   + Robot id
   + Time
   + Speed
   + Conditions of traffic (HEAVY, LIGHT, MODERATE). This could be a simple
     random choice.

Remarks
-------

 1. Assume that the robots follow a straight line between each point, travelling
    at constant speed.
 2. Disregard the fact that the start time is not in sync. The dispatcher can
    start pumping data as soon as it has read the files.
 3. A nearby station should be less than 350 meters from the robot's position.

Deliverables
------------

The assignment should be delivered as a command line program that allows the
user to start the simulation, and see (on the console or log file), the
communications between robots and dispatcher.

This is a fairly open assignment, in terms of how you structure the solution.
You will be judged on the overall quality of the code (simplicity,
presentation, performance).

The task should take no more than 4 hours to do. Your solution, as well as more
complex ideas, improvements, and suggestions will be discussed at the interview
stage.
