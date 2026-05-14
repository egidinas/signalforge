# Legacy Harvest Register

This register tracks ideas harvested from internal legacy repositories. It is
not evidence that source code, history, fixtures, or private context are safe to
copy.

| Idea | Candidate destination | Public-safe rule |
| --- | --- | --- |
| Visual tile pyramid | SignalForge graph/tile contracts | Keep as view-serving interfaces and fictional fixtures. |
| Capture classes | SignalForge capture policy contracts | Use generic signal classes only. |
| Campaign sealing | SignalForge archive/replay interfaces | Exclude private storage targets and procedures. |
| Signal aliases | SignalForge catalogue contracts | Caller provides aliases; no project defaults. |
| Passive bus monitoring | SignalForge transport interface candidate | Public API only; vendor ownership details stay downstream. |
| Reduction levels | SignalForge capture vocabulary | Keep generic levels; exclude private derived-channel names. |
| Protocol multiplexer | protocol-specific public library | Protocol-general implementation belongs outside SignalForge. |
| Operator confirmation flows | downstream private applications | Public demos may mock authority but must not claim live control proof. |
| Stability monitor core | SignalForge stability package | Keep generic thresholds, windows, and observations only; no private operator policy. |
| CAN trace framing and DLC helpers | SignalForge CAN trace package | Keep raw frame/PCAP helpers generic; vendor handles and bus topology stay downstream. |
| DBC metadata grouping | SignalForge DBC metadata package | Use public fixtures only; private DBC files and route defaults stay downstream. |
| Priority poll queues | SignalForge pollqueue package | Generic scheduling only; MeCom parameters and TEC semantics stay in Meerstetter-Go. |
| Ring buffer and dedup mechanics | SignalForge ringbuf package | Generic records and overwrite semantics only; device ring commands stay downstream. |
| Streaming reduction statistics | SignalForge stats package | Generic min/max/mean/stddev and windowing only; signal-specific policy stays downstream. |
| Passive transition/PID observation | SignalForge control-observation package | Recommendation primitives only; device write authority and safety policy stay downstream. |
| TDMS import contract | Loom importer or third-party library | SignalForge may own neutral import contracts, not private samples or lab procedures. |
