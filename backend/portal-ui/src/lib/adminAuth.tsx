import { createContext, useContext, useEffect, useState, ReactNode } from "react";
import { adminApi, AdminProfile } from "../api/admin";

interface AdminAuthCtx {
  me: AdminProfile | null;
  loading: boolean;
  refresh: () => Promise<void>;
  setMe: (m: AdminProfile | null) => void;
}

const Ctx = createContext<AdminAuthCtx>({
  me: null,
  loading: true,
  refresh: async () => {},
  setMe: () => {},
});

export function AdminAuthProvider({ children }: { children: ReactNode }) {
  const [me, setMe] = useState<AdminProfile | null>(null);
  const [loading, setLoading] = useState(true);

  const refresh = async () => {
    try {
      setMe(await adminApi.me());
    } catch {
      setMe(null);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { refresh(); }, []);

  return <Ctx.Provider value={{ me, loading, refresh, setMe }}>{children}</Ctx.Provider>;
}

export function useAdminAuth() {
  return useContext(Ctx);
}
