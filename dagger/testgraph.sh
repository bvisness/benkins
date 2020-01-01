#!/bin/bash
go test . -v && (dot -Tpng dotout > graph.png) && open graph.png
