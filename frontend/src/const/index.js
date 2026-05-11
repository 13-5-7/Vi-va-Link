export const ROLES = {
  OPERATOR: 'bus_operator',
  SHIPPER: 'shipper',
};

const BASE_API = '/api/v1';

export const API_PATH = {
  BASE: BASE_API,
  LOGIN: `${BASE_API}/auth/login`,
  LOGOUT: `${BASE_API}/auth/logout`,
  SCHEDULES_SEARCH: `${BASE_API}/schedules/search`,
  BOOKINGS: `${BASE_API}/bookings`,
  SCHEDULES: `${BASE_API}/schedules`,
  TRACKING: `${BASE_API}/tracking`,
  COMPANIES: `${BASE_API}/companies`,
  COMPANIES_ME: `${BASE_API}/companies/me`,
  COMPANIES_ME_STORAGE: `${BASE_API}/companies/me/storage`,
  AUTH_REGISTER: `${BASE_API}/auth/register`,
};