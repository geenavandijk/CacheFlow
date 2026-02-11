export const apiRequest = async (endpoint, options = {}) => {
  const headers = {
    'Content-Type': 'application/json',
    'x-cf-device-id': localStorage.getItem('x-cf-device-id'),
    'x-cf-uid': localStorage.getItem('x-cf-uid'),
    'x-cf-bearer': localStorage.getItem('x-cf-bearer'),
    'x-cf-refresh': localStorage.getItem('x-cf-refresh'),
    ...options.headers
  };

  return fetch(endpoint, {
    ...options,
    headers
  });
};