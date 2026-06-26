import { useSearchParams } from "react-router-dom";

export type ViewState = "default" | "loading" | "empty" | "error" | "denied" | "success" | "long";

export function useViewState(): ViewState {
  const [params] = useSearchParams();
  const state = params.get("state");
  if (state === "loading" || state === "empty" || state === "error" || state === "denied" || state === "success" || state === "long") {
    return state;
  }
  return "default";
}

