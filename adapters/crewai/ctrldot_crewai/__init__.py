"""
Ctrl Dot CrewAI Adapter

Guard LLM calls and tool execution in CrewAI projects.
"""

from .client import CtrlDotClient
from .guard import build_llm_call_proposal, build_tool_call_proposal
from .llm import CtrlDotLLM
from .tool import CtrlDotToolWrapper

__version__ = "0.1.0"
__all__ = [
    "CtrlDotClient",
    "CtrlDotLLM",
    "CtrlDotToolWrapper",
    "build_llm_call_proposal",
    "build_tool_call_proposal",
]
