# CalenAI Agent Reliability Roadmap

## ✅ 1. Tool Recursion / Infinite Loop Protection
Status: DONE

Problem:
- Agent could enter infinite tool-calling loops:
  LLM → Tool → LLM → Tool → ...

Risks:
- Stack overflow
- High CPU usage
- Hanging requests
- No response to user

Implemented:
- Added maxToolCalls limit
- Added depth tracking
- Stops execution when limit is reached

---

## 🔄 2. Multi-Tool Message Handling Cleanup
Status: PENDING

Problem:
- Assistant message is duplicated when multiple tool calls are returned in a single response.

Goal:
- Append assistant tool-call message once.
- Append all tool results after it.

Benefits:
- Cleaner conversation history
- Better model reasoning
- Easier debugging

---

## 🔄 3. Bulk Task Filtering System
Status: PENDING

Problem:
- Current deletion requires task IDs.
- Natural language deletion is difficult.

Examples:
- Delete all tasks this month
- Delete gym reminders
- Delete tomorrow's events

Planned Tools:
- count_tasks_by_filter
- delete_tasks_by_filter

Supported Filters:
- Date range
- Title contains
- Completion status

---

## 🔄 4. Safe Deletion Confirmation Workflow
Status: PENDING

Problem:
- Destructive actions execute immediately.

Desired Flow:

User:
Delete all tasks this month

AI:
Found 23 matching tasks.
Do you want me to delete them?

User:
Yes

AI:
Deleted 23 tasks.

Benefits:
- Prevents accidental mass deletion
- Improves trust and safety

---

## 🔄 5. Advanced Task Search
Status: PENDING

Goal:
Support natural language task lookup.

Examples:
- What tasks do I have tomorrow?
- Show GRE tasks
- When is my dentist appointment?

Planned Tool:
- find_tasks_by_filter

Benefits:
- Faster lookups
- Better context retrieval
- Reduced unnecessary get_tasks calls

---

## 🔄 6. User Ownership & Security Layer
Status: FUTURE

Goal:
Ensure users can only access their own tasks.

Implementation:
- Add UserID to Task model
- Filter all queries by UserID

Benefits:
- Multi-user support
- Security and privacy

---

## 🔄 7. Recovery / Undo System
Status: FUTURE

Goal:
Allow task restoration after deletion.

Implementation:
- Use GORM soft deletes
- Restore deleted tasks
- Undo last deletion

Benefits:
- Data recovery
- Reduced risk of accidental loss

---

# Progress Summary

[✓] Tool Recursion / Infinite Loop Protection
[ ] Multi-Tool Message Handling Cleanup
[ ] Bulk Task Filtering System
[ ] Safe Deletion Confirmation Workflow
[ ] Advanced Task Search
[ ] User Ownership & Security Layer
[ ] Recovery / Undo System