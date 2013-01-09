filehasher
==========

About
-----

This package allows for asynchronous hashing of files: The hashing of certain
files can be requested at any time and you can wait for the result.

The hashing can be paused or stopped at any time.

Internally the overall performance is maximized, i.e. by hashing files from different
physical devices in parallel (TODO).

Motivation
----------

I started filehasher mainly for two reasons:

* I need the functionality in another project of mine, f2fshare
* I wanted to get used to Go's concurrency idioms by working on a relatively simple project

I am also going to request a code review to get the most from the project.

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
