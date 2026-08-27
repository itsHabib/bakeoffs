# 60-second demo

Run the single command in the README. Point at these lines as they appear:

> `oracle H2: summary=SUPPORTED  coverage="3/3 required claims supported"  gaps=[]  merge_authority=none`

“The frozen assurance report says every required claim is supported. It also
says it has no merge authority. Without a mandate, the effect gateway denies.”

> `root -> child review_candidate: signatures valid`
>
> `audience-signed review request: ALLOW receipt sha256:...`

“The synthetic root is bound to task 17 revision 4 and the exact H2 base, head,
and diff. It permits inspect and review. Its child is signed by the root
audience, narrows to review, expires earlier, decrements delegation depth, and
names the worker public key. That worker signs this exact review request, so the
gateway emits a deterministic allow receipt.”

> `validly signed child adds publish_candidate:`
>
> `  signatures-only mutant: ALLOW  <-- planted bug`
>
> `  mandate: DENY scope_inflation`

“These are the same correctly signed widened child and publish request. The
mutant skips only the child-action subset check. Signature verification alone
allows publish; production rejects the first broken law as scope inflation.”

Point at the printed receipt and common envelope.

“The receipt carries the verified ancestry, request digest, exact subject,
decision, and stable reason. Two runs produce identical bytes. The envelope
copies the oracle fields unchanged and reports only this entry’s attenuation
status. No effect occurred.”

End on the comparator line:

“A standard caveated capability plus a receipt can enforce this same law, so
the custom-format bet has not beaten its cheap alternative. What is proven is
the missing authority boundary: delegation restrictions must travel with the
credential and monotonically narrow. Offline signatures do not solve
revocation, theft, global replay, or exactly-once consumption.”
