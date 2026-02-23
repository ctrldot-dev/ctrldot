"""
Minimal CrewAI example with Ctrl Dot guards.

This example demonstrates:
1. Registering an agent with Ctrl Dot
2. Using CtrlDotLLM to guard LLM calls
3. Using CtrlDotToolWrapper to guard tool execution
4. Running a CrewAI crew with guards enabled

Prerequisites:
- Ctrl Dot daemon running on http://127.0.0.1:7777
- CrewAI installed: pip install crewai
"""

import os
import uuid
from crewai import Agent, Task, Crew
from crewai.tools import BaseTool

# Import Ctrl Dot adapter (works when run from repo root after: pip install -e adapters/crewai)
try:
    from ctrldot_crewai import CtrlDotClient, CtrlDotLLM, CtrlDotToolWrapper
except ImportError:
    import sys
    _adapter_dir = os.path.join(os.path.dirname(__file__), "..")
    if _adapter_dir not in sys.path:
        sys.path.insert(0, _adapter_dir)
    from ctrldot_crewai import CtrlDotClient, CtrlDotLLM, CtrlDotToolWrapper


# Simple example tool
class HelloTool(BaseTool):
    name: str = "hello"
    description: str = "Says hello to someone"
    
    def _run(self, name: str) -> str:
        return f"Hello, {name}!"


def main():
    # Initialize Ctrl Dot client
    client = CtrlDotClient()

    # Pre-flight: ensure daemon is reachable (avoid confusing "connection closed" later)
    try:
        r = client.session.get(f"{client.base_url}/v1/health", timeout=5)
        r.raise_for_status()
    except Exception as e:
        print(f"❌ Cannot reach Ctrl Dot daemon at {client.base_url}")
        print(f"   Error: {e}")
        print("   Start the daemon in another terminal: DB_URL=... PORT=7777 ./bin/ctrldotd")
        raise SystemExit(1) from e

    # Agent and session IDs
    agent_id = os.getenv("CTRLDOT_AGENT_ID", "crewai-agent")
    agent_name = os.getenv("CTRLDOT_AGENT_NAME", "CrewAI Agent")
    session_id = os.getenv("CTRLDOT_SESSION_ID", f"session-{uuid.uuid4()}")
    
    print(f"🔐 Registering agent: {agent_id}")
    try:
        client.register_agent(agent_id, agent_name)
        print(f"✅ Agent registered")
    except Exception as e:
        print(f"⚠️  Registration warning (may already exist): {e}")
    
    # Create a simple LLM (you'll need to configure with your API key)
    # For this example, we'll use a mock/dummy LLM
    class DummyLLM:
        def call(self, messages, **kwargs):
            if isinstance(messages, str):
                return f"Mock response to: {messages}"
            return "Mock response"
    
    dummy_llm = DummyLLM()
    
    # Wrap LLM with Ctrl Dot guard
    print(f"🛡️  Creating guarded LLM")
    guarded_llm = CtrlDotLLM(
        model="claude-opus-4.6",
        provider_llm=dummy_llm,
        ctrldot_client=client,
        agent_id=agent_id,
        session_id=session_id,
        cheap_model="claude-sonnet-4"
    )
    
    # Create tool and wrap it (use .wrap() so CrewAI Agent accepts it as BaseTool)
    hello_tool = HelloTool()
    guarded_tool = CtrlDotToolWrapper.wrap(
        hello_tool, client, agent_id, session_id
    )
    
    # Create CrewAI agent
    print(f"🤖 Creating CrewAI agent")
    agent = Agent(
        role="Assistant",
        goal="Help users",
        backstory="A helpful assistant",
        llm=guarded_llm,
        tools=[guarded_tool],
        verbose=True
    )
    
    # Create task
    task = Task(
        description="Say hello to Alice",
        agent=agent,
        expected_output="A greeting message"
    )
    
    # Create crew
    crew = Crew(
        agents=[agent],
        tasks=[task],
        verbose=True
    )
    
    # Run crew
    print(f"🚀 Running crew...")
    try:
        result = crew.kickoff()
        print(f"✅ Crew completed: {result}")
    except Exception as e:
        print(f"❌ Crew failed: {e}")
        raise


if __name__ == "__main__":
    main()
