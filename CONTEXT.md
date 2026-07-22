# Domain language

## Application

A deployer-owned unit whose desired state, synchronization, health, and delivery progress are considered together. An Application is the primary object deployers inspect and act upon.

## Delivery controller

The mechanism that brings an Application's workload revision into service. A delivery controller may perform a standard rolling deployment or a progressive rollout and is subordinate to its Application.

## Promotion

An explicit deployer decision that allows a paused progressive delivery to continue.

## Advanced resource

A cluster-level implementation resource exposed for diagnosis rather than routine deployment work. Advanced resources do not define the primary deployer-facing organization of the system.
