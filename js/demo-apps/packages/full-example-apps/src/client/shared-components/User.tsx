import {Card, CardContent, CardHeader, Grid, Typography, Button, FormLabel} from '@material-ui/core';
import React, {useEffect, useState} from 'react';
import {
  loadCurrentUserAPI,
  performSaveUserPropertiesAPI
} from "../dedicated-tokenizer/passport-auth-example/utils/api-facade";
import {UserModel} from "../../shared/types";
import {lunaSecDomain} from "../constants";
import {LunaSecConfigContext, SecureForm, SecureInput} from "@lunasec/react-sdk";

interface SecureUserPropertiesProps {
  ssnToken: string
  setSSNToken: React.Dispatch<string>
}

async function loadUser(setUser, setError) {
  const currentUserResp = await loadCurrentUserAPI();
  if (currentUserResp.success) {
    setUser(currentUserResp.user);
    return;
  }
  setError(currentUserResp.error)
}

export const User: React.FunctionComponent = () => {
  const [authError, setAuthError] = useState<string>('')
  const [saveSuccessful, setSaveSuccessful] = useState<boolean>(null);
  const [error, setError] = useState<string>(null);
  const [user, setUser] = useState<UserModel>(null);

  const [ssnToken, setSSNToken] = useState<string>(null);

  const ssnValidated = (isValid) => {
    console.log(`ssn is valid? ${isValid}`)
  }

  const persistTokens = (e) => {
    e.preventDefault();
    setSSNToken(e.target.value);
  }

  useEffect(() => {
    void loadUser(setUser, setError);
  }, []);

  const saveProperties = async () => {
    if (ssnToken === null) {
      setError('ssnToken is null');
      return;
    }

    const res = await performSaveUserPropertiesAPI({
      ssnToken: ssnToken
    });
    if (!res.success) {
      setError(JSON.stringify(res.error));
      return;
    }
    setSaveSuccessful(true);
  }

  if (user === null) {
    return (
      <Grid item xs={12}>
        <Card>
          <CardContent><p>Loading...</p></CardContent>
        </Card>
      </Grid>
    );
  }

  return (
    <LunaSecConfigContext.Provider
      value={{
        lunaSecDomain: lunaSecDomain,
        authenticationErrorHandler: (_e: Error) => {
          setAuthError('Failed to authenticate with LunaSec. \n Is a user logged in?');
        },
      }}
    >
      {authError !== null ? (<p>{authError}</p>) : null}
      <Grid item xs={12}>
        {error !== null
        ? (<Card>
              <CardHeader title={'Error'} />
              <CardContent><p>{error}</p></CardContent>
            </Card>)
        : null}
        <Card>
          <CardHeader title={`User: ${user.username}`}/>
          <CardContent>
            <SecureForm name="secure-form-example" onSubmit={(e) => persistTokens(e)}>
              <Typography>Id: {user.id}</Typography>
              <FormLabel htmlFor="ssn-token-input">
                Social Security Number
              </FormLabel>
              <SecureInput
                id="ssn-token-input"
                name="ssn"
                type="ssn"
                validator="SSN"
                onValidate={(isValid) => ssnValidated(isValid)}
                token={ssnToken}
                placeholder="Enter Your SSN"
              />
              <Button
                variant="outlined"
                color="primary"
                style={{ textTransform: "none" }}
                type="submit"
              >
                Save
              </Button>
            </SecureForm>
          </CardContent>
        </Card>
      </Grid>
    </LunaSecConfigContext.Provider>
  );
};