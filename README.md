filehasher
==========

About
-----

This package allows for asynchronous hashing of files.

Goals
-----

These are the original goals of the project:

* Learn about Go and its concurrency idioms by working on a simple software project that involves parallel computing.
* Ideally, the resulting software can be reused in other projects.

Requirements
------------

* the library/application is for hashing files (e.g. SHA1).
* the interface should support that the caller specify the type of hash function.
* the application should be able to do I/O and hashing in parallel to optimize the overall performance.
* it should be possible to pause/resume the hashing at any time.
* it should also be possible to shutdown the concurrent go routines completely at any time (not just pause them).
* later it should be possible to implement advanced optimizations like hashing files from different physical devices in parallel.
* it should be possible to queue files for hashing at any time.

State
-----

The exported interface is not yet stable. Major changes are to be expected.

License
-------

Note: If you need a more permissive license, feel free to contact me.

> Copyright (C) 2012-2013 Robert Nitsch
> 
> This program is free software: you can redistribute it and/or modify
> it under the terms of the GNU General Public License as published by
> the Free Software Foundation, either version 3 of the License, or
> (at your option) any later version.
> 
> This program is distributed in the hope that it will be useful,
> but WITHOUT ANY WARRANTY; without even the implied warranty of
> MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
> GNU General Public License for more details.
> 
> You should have received a copy of the GNU General Public License
> along with this program.  If not, see <http://www.gnu.org/licenses/>.
