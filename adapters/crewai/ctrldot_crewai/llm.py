"""Guarded LLM wrapper for CrewAI."""

import logging
from typing import List, Dict, Any, Optional, Union, TYPE_CHECKING

from .client import CtrlDotClient, CtrlDotDeniedError
from .guard import build_llm_call_proposal

if TYPE_CHECKING:
    from pydantic import BaseModel
    from crewai.agent.core import Agent
    from crewai.task import Task

logger = logging.getLogger(__name__)


def _create_ctrldot_llm_class():
    """Build a BaseLLM subclass so CrewAI's create_llm accepts it."""
    from crewai.llms.base_llm import BaseLLM

    class CtrlDotLLM(BaseLLM):
        """BaseLLM subclass that guards LLM calls via Ctrl Dot."""

        def __init__(
            self,
            model: str,
            provider_llm: Any,
            ctrldot_client: CtrlDotClient,
            agent_id: str,
            session_id: str,
            cheap_model: Optional[str] = None,
            **kwargs
        ):
            super().__init__(model=model, **kwargs)
            self.provider_llm = provider_llm
            self.client = ctrldot_client
            self.agent_id = agent_id
            self.session_id = session_id
            self.cheap_model = cheap_model
            self.provider_kwargs = kwargs

        def call(
            self,
            messages: Union[str, List[Dict[str, str]]],
            tools: Optional[List[dict]] = None,
            callbacks: Optional[List[Any]] = None,
            available_functions: Optional[Dict[str, Any]] = None,
            from_task: Any = None,
            from_agent: Any = None,
            response_model: Any = None,
            **kwargs
        ) -> Union[str, Any]:
            if isinstance(messages, str):
                messages_list = [{"role": "user", "content": messages}]
            else:
                messages_list = messages
            proposal = build_llm_call_proposal(
                agent_id=self.agent_id,
                session_id=self.session_id,
                model=self.model,
                messages=messages_list,
                est_tokens=kwargs.get("max_tokens"),
                est_gbp=None,
                meta={"tools_count": len(tools) if tools else 0}
            )
            try:
                decision = self.client.propose_action(proposal)
            except CtrlDotDeniedError as e:
                logger.error(f"Ctrl Dot denied LLM call: {e.reason}")
                raise
            if decision.get("decision") == "THROTTLE":
                logger.warning(f"Ctrl Dot throttling LLM call: {decision.get('reason', '')}")
                throttle_kwargs = kwargs.copy()
                if decision.get("model_policy") == "cheap" and self.cheap_model:
                    throttle_kwargs["max_tokens"] = min(throttle_kwargs.get("max_tokens", 4096), 2048)
                    throttle_kwargs["temperature"] = min(throttle_kwargs.get("temperature", 1.0), 0.7)
                else:
                    throttle_kwargs["max_tokens"] = int(throttle_kwargs.get("max_tokens", 4096) * 0.7) if "max_tokens" in throttle_kwargs else 2048
                    throttle_kwargs["temperature"] = (throttle_kwargs.get("temperature", 1.0) * 0.8) if "temperature" in throttle_kwargs else 0.7
                kwargs = throttle_kwargs
            if decision.get("decision") == "WARN":
                logger.warning(f"Ctrl Dot warning: {decision.get('reason', '')}")
            try:
                return self.provider_llm.call(
                    messages=messages,
                    tools=tools,
                    callbacks=callbacks,
                    available_functions=available_functions,
                    **kwargs
                )
            except Exception as e:
                logger.error(f"Provider LLM call failed: {e}")
                raise

        def __getattr__(self, name: str):
            return getattr(self.provider_llm, name)

    return CtrlDotLLM


# Public class: instantiate via the BaseLLM subclass
CtrlDotLLM = _create_ctrldot_llm_class()
