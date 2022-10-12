# Func record

1 Listen on port 8848, function realized under path /url
2 If access to other path, will reply no function under this path
3 path url without query will reply aks result
4 if query added, accepted format is "url=http(s)://xxx", format invalid will be rejected.
5 if query address is not reachable, will throughout the error directly to user(not safe 😋)
