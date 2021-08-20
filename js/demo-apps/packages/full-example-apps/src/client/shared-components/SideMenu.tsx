import { Drawer, List, ListItem, ListItemIcon, ListItemText, makeStyles } from '@material-ui/core';
import { createStyles, Theme } from '@material-ui/core/styles';
import React from 'react';
import { NavLink } from 'react-router-dom';
import {VpnKey, LockOpen, Home, AccountCircle, Folder} from "@material-ui/icons";

const drawerWidth = 240;

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    drawer: {
      width: drawerWidth,
      flexShrink: 0,
    },
    drawerPaper: {
      width: drawerWidth,
    },
    toolbar: theme.mixins.toolbar,
  }),
);

export const SideMenu: React.FunctionComponent = () => {
  const classes = useStyles({});
  return (
    <Drawer
      className={classes.drawer}
      variant='permanent'
      classes={{
        paper: classes.drawerPaper,
      }}
    >
      <div className={classes.toolbar} />
      <List>
        <ListItem button component={NavLink} to='/'>
          <ListItemIcon>
            <Home />
          </ListItemIcon>
          <ListItemText primary='Home' />
        </ListItem>
        <ListItem button component={NavLink} to='/signup'>
          <ListItemIcon>
            <VpnKey />
          </ListItemIcon>
          <ListItemText primary='Signup' />
        </ListItem>
        <ListItem button component={NavLink} to='/login'>
          <ListItemIcon>
            <LockOpen />
          </ListItemIcon>
          <ListItemText primary='Login' />
        </ListItem>
        <ListItem button component={NavLink} to='/user'>
            <ListItemIcon>
              <AccountCircle />
            </ListItemIcon>
            <ListItemText primary='User' />
        </ListItem>
        <ListItem button component={NavLink} to='/documents'>
          <ListItemIcon>
            <Folder />
          </ListItemIcon>
          <ListItemText primary='Documents' />
        </ListItem>
      </List>
    </Drawer>
  );
};