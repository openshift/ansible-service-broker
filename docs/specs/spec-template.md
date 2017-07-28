# Example spec

## Introduction
Introductory paragraph to your spec. What is the goal of the spec?

## Problem Description
Describe the problem you're trying to solve. Then answer how you're going
to solve it in the <Implementation Details>.

## <Implementation Details>
The body of the spec will be filled with details about the feature you're
trying to land.

Some good questions to answer in detail:
 - How will this improve the broker?
 - How will the broker's behavior change?
 - Will this change APBs?

## Work Items
A list of items that you plan to implement. You don't have to follow it
exactly during implementation, but it's good to compact all the details
about the spec into a series of steps that anyone can follow.

example:
 - Add a new pkg SpeedUpBindings
 - Build SpeedUpBindings functions so bindings are faster
 - Integrate SpeedUpBindings into the Binding workflow
 - Fix tests so they use SpeedUpBindings
