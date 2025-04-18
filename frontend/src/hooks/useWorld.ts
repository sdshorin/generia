import { useContext } from 'react';
import { WorldContext } from '../context/WorldContext';

export const useWorld = () => {
  return useContext(WorldContext);
};