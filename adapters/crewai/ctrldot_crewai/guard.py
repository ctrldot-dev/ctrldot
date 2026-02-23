"""Helpers to build Ctrl Dot proposal payloads."""

from typing import Dict, Any, Optional, List


def build_llm_call_proposal(
    agent_id: str,
    session_id: str,
    model: str,
    messages: List[Dict[str, Any]],
    est_tokens: Optional[int] = None,
    est_gbp: Optional[float] = None,
    meta: Optional[Dict[str, Any]] = None
) -> Dict[str, Any]:
    """
    Build a proposal payload for an LLM call.
    
    Args:
        agent_id: Agent identifier
        session_id: Session identifier
        model: Model name (e.g., "claude-opus-4.6")
        messages: List of message dicts
        est_tokens: Estimated token count (if None, naive estimate from text length)
        est_gbp: Estimated cost in GBP (if None, defaults to 0)
        meta: Optional metadata dict
        
    Returns:
        Proposal payload dict
    """
    # Naive token estimation if not provided
    if est_tokens is None:
        text = " ".join(str(m.get("content", "")) for m in messages)
        est_tokens = len(text) // 4  # Rough estimate
    
    if est_gbp is None:
        est_gbp = 0.0
    
    return {
        "agent_id": agent_id,
        "session_id": session_id,
        "intent": {
            "title": f"LLM call: {model}"
        },
        "action": {
            "type": "llm.call",
            "target": {
                "model": model,
                "message_count": len(messages)
            },
            "inputs": {
                "messages": messages
            }
        },
        "cost": {
            "currency": "GBP",
            "estimated_gbp": est_gbp,
            "estimated_tokens": est_tokens,
            "model": model
        },
        "context": {
            "tool": "llm",
            "tags": ["llm", "ai"],
            "meta": meta or {}
        }
    }


def build_tool_call_proposal(
    agent_id: str,
    session_id: str,
    tool_name: str,
    args: Dict[str, Any],
    est_tokens: Optional[int] = None,
    est_gbp: Optional[float] = None,
    meta: Optional[Dict[str, Any]] = None
) -> Dict[str, Any]:
    """
    Build a proposal payload for a tool call.
    
    Args:
        agent_id: Agent identifier
        session_id: Session identifier
        tool_name: Tool name (e.g., "git.push", "filesystem.write")
        args: Tool arguments dict
        est_tokens: Estimated token count (optional)
        est_gbp: Estimated cost in GBP (optional, defaults to 0)
        meta: Optional metadata dict
        
    Returns:
        Proposal payload dict
    """
    if est_tokens is None:
        est_tokens = 0
    
    if est_gbp is None:
        est_gbp = 0.0
    
    return {
        "agent_id": agent_id,
        "session_id": session_id,
        "intent": {
            "title": f"Tool call: {tool_name}"
        },
        "action": {
            "type": f"tool.call.{tool_name}",
            "target": args,
            "inputs": {}
        },
        "cost": {
            "currency": "GBP",
            "estimated_gbp": est_gbp,
            "estimated_tokens": est_tokens,
            "model": "tool-execution"
        },
        "context": {
            "tool": tool_name,
            "tags": ["tool", tool_name],
            "meta": meta or {}
        }
    }
