import { createContext, useContext, useEffect, useState, ReactNode } from "react";
import { api, CustomerProfile } from "../api/client";

interface AuthCtx {
  me: CustomerProfile | null;
  loading: boolean;
  refresh: () => Promise<void>;
  setMe: (m: CustomerProfile | null) => void;
}

const Ctx = createContext<AuthCtx>({
  me: null,
  loading: true,
  refresh: async () => {},
  setMe: () => {},
});

export function AuthProvider({ children }: { children: ReactNode }) {
  const [me, setMe] = useState<CustomerProfile | null>(null);
  const [loading, setLoading] = useState(true);

  const refresh = async () => {
    try {
      const m = await api.me();
      setMe(m);
    } catch {
      setMe(null);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    refresh();
  }, []);

  return <Ctx.Provider value={{ me, loading, refresh, setMe }}>{children}</Ctx.Provider>;
}

export function useAuth() {
  return useContext(Ctx);
}
