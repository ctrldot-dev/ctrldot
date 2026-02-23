"""Guarded tool wrapper for CrewAI."""

import logging
from typing import Any, Dict, Optional, TYPE_CHECKING

from .client import CtrlDotClient, CtrlDotDeniedError
from .guard import build_tool_call_proposal

if TYPE_CHECKING:
    from crewai.tools import BaseTool

logger = logging.getLogger(__name__)


def _make_ctrldot_tool_wrapper(
    inner: Any,
    ctrldot_client: CtrlDotClient,
    agent_id: str,
    session_id: str,
) -> Any:
    """
    Wrap a CrewAI BaseTool with Ctrl Dot guard. Returns a BaseTool subclass
    so CrewAI Agent accepts it (Pydantic model_type).
    """
    from crewai.tools import BaseTool
    from pydantic import Field

    tool_name = getattr(inner, "name", "unknown_tool")
    tool_description = getattr(inner, "description", "Guarded by Ctrl Dot") or "Guarded by Ctrl Dot"
    args_schema = getattr(inner, "args_schema", None)

    class _GuardedTool(BaseTool):
        name: str = Field(default=tool_name, description=tool_description)
        description: str = Field(default=tool_description, description="Tool description")

        def _run(self, **kwargs) -> str:
            return _run_guarded(
                inner=inner,
                client=ctrldot_client,
                agent_id=agent_id,
                session_id=session_id,
                tool_name=tool_name,
                kwargs=kwargs,
            )

    if args_schema is not None and getattr(BaseTool, "_ArgsSchemaPlaceholder") != args_schema:
        _GuardedTool.args_schema = args_schema
    _GuardedTool.__name__ = f"Guarded_{getattr(inner.__class__, '__name__', 'Tool')}"
    return _GuardedTool(name=tool_name, description=tool_description)


def _run_guarded(
    inner: Any,
    client: CtrlDotClient,
    agent_id: str,
    session_id: str,
    tool_name: str,
    kwargs: Dict[str, Any],
) -> str:
    """Shared guard logic used by wrapper."""
    proposal = build_tool_call_proposal(
        agent_id=agent_id,
        session_id=session_id,
        tool_name=tool_name,
        args=kwargs,
        est_tokens=0,
        est_gbp=0.0,
        meta={"tool_class": type(inner).__name__},
    )
    try:
        decision = client.propose_action(proposal)
    except CtrlDotDeniedError as e:
        logger.error(f"Ctrl Dot denied tool call {tool_name}: {e.reason}")
        raise
    if decision.get("decision") == "WARN":
        logger.warning(f"Ctrl Dot warning for {tool_name}: {decision.get('reason', '')}")
    if decision.get("decision") == "THROTTLE":
        logger.warning(f"Ctrl Dot throttling {tool_name}: {decision.get('reason', '')}")
    if hasattr(inner, "_run"):
        result = inner._run(**kwargs)
    elif hasattr(inner, "run"):
        result = inner.run(**kwargs)
    elif callable(inner):
        result = inner(**kwargs)
    else:
        raise ValueError(f"Tool {tool_name} has no callable run method")
    return str(result)


class CtrlDotToolWrapper:
    """
    Wraps a CrewAI BaseTool to guard tool execution via Ctrl Dot.
    Use wrap_tool() for CrewAI Agent compatibility (returns a BaseTool subclass).
    
    Usage:
        from crewai.tools import BaseTool
        from ctrldot_crewai import CtrlDotToolWrapper, CtrlDotClient
        
        client = CtrlDotClient()
        inner_tool = SomeTool()
        guarded_tool = CtrlDotToolWrapper.wrap(
            inner_tool, client, "my-agent", "session-123"
        )
    """

    @staticmethod
    def wrap(
        inner: Any,
        ctrldot_client: CtrlDotClient,
        agent_id: str,
        session_id: str,
    ) -> Any:
        """Wrap a BaseTool for use with CrewAI Agent. Returns a BaseTool-compatible instance."""
        return _make_ctrldot_tool_wrapper(inner, ctrldot_client, agent_id, session_id)

    def __init__(
        self,
        inner: Any,  # BaseTool from CrewAI
        ctrldot_client: CtrlDotClient,
        agent_id: str,
        session_id: str
    ):
        """
        Initialize guarded tool wrapper.
        Prefer CtrlDotToolWrapper.wrap() for use with CrewAI Agent.
        """
        self.inner = inner
        self.client = ctrldot_client
        self.agent_id = agent_id
        self.session_id = session_id

    def _run(self, **kwargs) -> str:
        """
        Run tool with Ctrl Dot guard.
        
        Args:
            **kwargs: Tool arguments
            
        Returns:
            Tool output string
            
        Raises:
            CtrlDotDeniedError: If Ctrl Dot denies the tool call
        """
        # Get tool name (try various attributes)
        tool_name = getattr(self.inner, "name", None) or \
                   getattr(self.inner, "__class__", {}).__name__ if hasattr(self.inner, "__class__") else \
                   "unknown_tool"
        
        # Build proposal
        proposal = build_tool_call_proposal(
            agent_id=self.agent_id,
            session_id=self.session_id,
            tool_name=tool_name,
            args=kwargs,
            est_tokens=0,
            est_gbp=0.0,
            meta={"tool_class": type(self.inner).__name__}
        )
        
        # Propose action
        try:
            decision = self.client.propose_action(proposal)
        except CtrlDotDeniedError as e:
            logger.error(f"Ctrl Dot denied tool call {tool_name}: {e.reason}")
            raise
        
        # Handle warn mode
        if decision.get("decision") == "WARN":
            logger.warning(f"Ctrl Dot warning for {tool_name}: {decision.get('reason', '')}")
        
        # Handle throttle mode
        if decision.get("decision") == "THROTTLE":
            logger.warning(f"Ctrl Dot throttling {tool_name}: {decision.get('reason', '')}")
            # Still allow, but could add rate limiting here
        
        # Run inner tool
        try:
            if hasattr(self.inner, "_run"):
                result = self.inner._run(**kwargs)
            elif hasattr(self.inner, "run"):
                result = self.inner.run(**kwargs)
            elif callable(self.inner):
                result = self.inner(**kwargs)
            else:
                raise ValueError(f"Tool {tool_name} has no callable run method")
            
            return str(result)
        except Exception as e:
            logger.error(f"Tool {tool_name} execution failed: {e}")
            raise
    
    def __getattr__(self, name: str):
        """Delegate other attributes to inner tool."""
        return getattr(self.inner, name)
