Pings a host to check network connectivity and response time.

<usage>
- Provide host name or IP address
- Sends ICMP echo requests
- Returns response time and status
- Tests network reachability
</usage>

<features>
- ICMP ping
- Response time measurement
- Packet loss detection
- Multiple attempts
- DNS resolution
</features>

<prerequisites>
1. Network connectivity
2. ICMP not blocked by firewall
</prerequisites>

<parameters>
1. host: Host to ping (required)
</parameters>

<special_cases>
- Unreachable host: Timeout
- IPv6: Supported when available
- Localhost: Always works
- Firewalls: May block ICMP
</special_cases>

<critical_requirements>
- Valid hostname or IP
- Network interface available
- ICMP not blocked
</critical_requirements>

<warnings>
- Some networks block ICMP
- Firewalls may interfere
- Rate limiting possible
- IPv4 vs IPv6 differences
</warnings>

<recovery_steps>
If ping fails:
1. Check network connection
2. Verify hostname spelling
3. Try IP address instead
4. Check firewall settings
</recovery_steps>

<best_practices>
- Ping multiple times for accuracy
- Check DNS resolution first
- Use IP address to isolate DNS issues
</best_practices>

<examples>
✅ Correct: Ping website
```
system_net_ping(host="example.com")
```

✅ Correct: Ping IP address
```
system_net_ping(host="8.8.8.8")
```
</examples>

<cross_platform>
- Works on all platforms
- Permission requirements vary
- IPv6 support varies
</cross_platform>
