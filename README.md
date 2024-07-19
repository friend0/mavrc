# MavRC

MavRC is a simple implementation of a bridge between NATS and either a Mavlink, or RC Connection.
The goal is to be able to send control commands to the broadest range of vehicles possible.

## How does it work?
MavRC runs a NATS service, subscribing to messages on the "mavrc" topic.
"Remotes" are sinks for control commands. Concretely, a Remote is either a proxy to a Mavlink vehicle,
or a Serial link to a CyberTX instance. See (here)[github.com/friend0/CyberTX]. 

Mavlink is an industry standard way to communicate to/from open source autopilots like PX4, and Ardupilot, 
among others. 

CyberTX is a serial/protobuf to PPM converter that can drive trainer inputs to RC radio equipment.
In short, if you can control it with an RC, and your RC has trainer input, 
you can control that vehicle from a computer using CyberTX and MavRC.
