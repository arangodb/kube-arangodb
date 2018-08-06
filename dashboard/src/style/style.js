import styled from 'react-emotion';
import { Menu } from 'semantic-ui-react';

export const Field = styled('div')`
  padding-top: 0.3em;
  padding-bottom: 0.3em;
  clear: both;
`;

export const FieldLabel = styled('span')`
  width: 9rem;
  display: inline-block;
  float: left;
`;

export const FieldContent = styled('div')`
  display: inline-block;
`;

export const FieldIcons = styled('div')`
  float: right;
`;

export const LoaderBox = styled('span')`
  float: right;
  width: 0;
  padding-right: 1em;
  margin-right: 1em;
  margin-top: 1em;
  max-width: 0;
  display: inline-block;
`;

export const LoaderBoxForTable = styled('span')`
  float: right;
  width: 0;
  padding-right: 1em;
  max-width: 0;
  display: inline-block;
`;

export const StyledMenu = styled(Menu)`
  width: 15rem !important;
  @media (max-width: 768px) {
    width: 10rem !important;
  }
`;

export const StyledContentBox = styled('div')`
  margin-left: 15rem;
  @media (max-width: 768px) {
    margin-left: 10rem;
  }
`;
