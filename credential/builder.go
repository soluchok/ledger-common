package credential

import (
	"gopkg.in/go-playground/validator.v9"

	"github.com/workdaycredentials/ledger-common/proof"
)

// Builder is used to construct signed Verifiable Credential.
type Builder struct {
	// SubjectDID is the Decentralized ID of the subject of the credential, who is normally also the
	// credential Holder. This is recorded as the "id" (JSON-LD "@id") property in the
	// credentialSubject block of the credential.
	SubjectDID string `validate:"required"`
	// Data is a map of claims that adhere to the schema referenced in the Metadata.
	Data map[string]interface{}
	// Metadata is information about the credential.
	Metadata *Metadata `validate:"required"`
	// Signer has the ability to generate a digital signature for a provided signature type.
	Signer proof.Signer `validate:"required"`
	// SignatureType specifies the suite used to generate the credential signature
	SignatureType proof.SignatureType `validate:"required"`
}

// Build returns a signed Verifiable Credential using the current state of the builder.
func (b Builder) Build() (*VerifiableCredential, error) {
	if err := validator.New().Struct(b); err != nil {
		return nil, err
	}

	suite, err := proof.SignatureSuites().GetSuiteForCredentials(b.SignatureType, proof.V2)
	if err != nil {
		return nil, err
	}

	// The "id" attribute is added if missing from the claim data.
	var credSubjects = map[string]interface{}{SubjectIDAttribute: b.SubjectDID}
	for k, v := range b.Data {
		credSubjects[k] = v
	}

	options := &proof.ProofOptions{ProofPurpose: proof.AssertionMethodPurpose}
	// Compute the claim proofs for selective disclosure.
	var claimProofs = make(map[string]proof.Proof, len(credSubjects))
	for k, v := range credSubjects {
		credential := &VerifiableCredential{
			UnsignedVerifiableCredential: UnsignedVerifiableCredential{
				Metadata:          *b.Metadata,
				CredentialSubject: map[string]interface{}{k: v},
			},
		}
		if err := suite.Sign(credential, b.Signer, options); err != nil {
			return nil, err
		}
		claimProofs[k] = *credential.Proof
	}

	cred := &VerifiableCredential{
		UnsignedVerifiableCredential: UnsignedVerifiableCredential{
			Metadata:          *b.Metadata,
			CredentialSubject: credSubjects,
			ClaimProofs:       claimProofs,
		},
	}
	return cred, suite.Sign(cred, b.Signer, options)
}
