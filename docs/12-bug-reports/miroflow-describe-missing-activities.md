Bug Report: DESCRIBE MICROFLOW Returns Incomplete MDL Output
Severity: Medium
Component: mxcli — DESCRIBE MICROFLOW command
Affected Version: mxcli (MxSummit.mpr, Mendix 11.6.3)

Summary
DESCRIBE MICROFLOW produces truncated MDL output that does not accurately represent the full logic of a microflow. The generated MDL is missing activities, return statements, and IF block bodies, despite the catalog reporting significantly more activities.

Steps to Reproduce

echo "DESCRIBE MICROFLOW Administration.ChangePassword;" | ./mxcli -p MxSummit.mpr
echo "DESCRIBE MICROFLOW FeedbackModule.VAL_Feedback;" | ./mxcli -p MxSummit.mpr
Expected Behavior
The MDL output should reflect all activities present in the microflow, including:

Complete IF block bodies
RETURN statements for microflows with a declared return type
Java action calls and other activity types
Actual Behavior
Administration.ChangePassword (catalog: 10 activities)

The IF block body is empty — no password change logic is emitted
No RETURN statement, despite the microflow likely returning a result
FeedbackModule.VAL_Feedback (catalog: 28 activities)

Declares $ValidFeedback Boolean = true but never returns it
The IF block body is empty — no validation feedback or assignments are emitted
Only 1 of 28 catalogued activities is represented
Impact
Round-tripping (describe → edit → re-apply) is not possible since the output is incomplete
Developers cannot use DESCRIBE as a reliable source of truth for microflow logic
Misleading output may cause developers to believe microflows are simpler than they are
Additional Notes
The discrepancy appears correlated with activities that have no direct MDL equivalent (e.g. Java action calls, exclusive splits/merges, annotations). It is unclear whether these are silently dropped or simply not yet supported in the MDL emitter.
