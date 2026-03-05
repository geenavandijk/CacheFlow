export interface RiskSettings {
  budget?: number | null;
  max_loss_percentage?: number | null;
  risk_tolerance?: number | null;
}

export interface LoadInAccountData {
  account_id: string;
  email: string;
  first_name: string;
  last_name: string;
  risk_settings?: RiskSettings | null;
}