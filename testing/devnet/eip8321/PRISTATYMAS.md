# Augusto demonstracijos atmintinė

## Ką sakyti pradžioje

„Mūsų komanda įgyvendina EIP-8321 kaip atskirą eksperimentinį atnaujinimą po Heze. Jis leidžia RANDAO naudoti iš anksto užregistruotą maišos grandinę. Mano dalis apėmė naujus duomenų tipus, jų saugojimą ir sąsajas, o dabar – bendro kliento paleidimą vietiniame tinkle.“

## Ką rodome

1. Du mūsų Prysm mazgai ir jų validatoriai. Kiekvienas mazgas prijungtas prie savo Geth vykdymo kliento. Iš viso 128 testiniai validatoriai.
2. Tinklas pradeda veikti pagal Heze taisykles. Ketvirtoje epochoje aktyvuojamas atskiras EIP-8321 forkas.
3. Blokuose atsiranda pasirašytos RANDAO įsipareigojimų registracijos. Validatoriai išsaugo privačią grandinės pradžią savo duomenų bazėje.
4. Po aktyvavimo delsos bloko siūlytojas pateikia ankstesnę grandinės reikšmę. Mazgas patikrina jos maišą ir palygina su saugomu įsipareigojimu.
5. Abu mazgai sutaria dėl bloko ir toliau finalizuoja grandinę. Tai tikriname tiesiogiai jų API, o ne sprendžiame vien iš testo „PASS“.

## Pagrindinė formulė

`H(Y) = X`, kur `H(Y) = BLAKE3("HASH_CHAIN_RANDAO" || Y)`.

- `X` – anksčiau saugotas įsipareigojimas.
- `Y` – dabar atskleidžiama 32 baitų reikšmė.
- Patikrinus, saugomas įsipareigojimas pakeičiamas į `Y`.
- Šios reikšmės yra maišos grandinės nariai. Ethereum blokas tik perneša atskleidimą.

Kol validatorius neturi aktyvaus įsipareigojimo, naudojamas ankstesnis BLS RANDAO kelias. Įregistravus ir aktyvavus įsipareigojimą, bloke naudojamas `hash_chain_reveal`, o `randao_reveal` turi specialią BLS begalybės taško reikšmę.

## Komandos demonstracijai

Prysm kataloge:

```powershell
./testing/devnet/eip8321/Start.ps1
./testing/devnet/eip8321/Status.ps1
./testing/devnet/eip8321/Start-Dashboard.ps1
./testing/devnet/eip8321/Collect-Evidence.ps1
```

`Start.ps1` skirtas jau sukurtiems atvaizdams. Pirmą kartą paleisti `Build.ps1`. Demonstracijai atidaryti `Start-Dashboard.ps1` išspausdintą adresą: peržiūra rodo tikrus abiejų mazgų duomenis ir RANDAO laukus. Dora tinka Heze daliai; jos dabartinė versija dar nemoka naujo EIP-8321 bloko formato. `Collect-Evidence.ps1` išsaugo blokus ir patikrų suvestinę. Iš tos pačios medžiagos `verify_hash_chain.py` nepriklausomai patikrina BLAKE3 grandinės ryšius, nereikalaudamas privačių raktų.

Šiame mažame tinkle epocha turi 8 slotus po 6 sekundes. Įprastas pagrindinio Ethereum tinklo epochų dydis skiriasi. Ketvirta epocha prasideda ties 32 slotu; registracijų aktyvavimo delsa taip pat matuojama epochomis.

## Jei dėstytojas klausia apie testus

„Skiriame vienetinius testus, specifikacijos vektorius ir gyvo tinklo bandymą. Vienetinių testų ir vektorių rezultatai taikomi konkrečioms tikrintoms paketų grupėms; neteigiame, kad patikrinome visą Prysm ar pasiekėme visų klientų suderinamumą. Šioje demonstracijoje papildomai rodome tikrus blokus, registracijas, atskleidimus, mazgų sutarimą ir finalizaciją po forko.“

Galutinius skaičius imti iš konkretaus bandymo `summary.json` ir testų ataskaitos. Dora palaikymas ir tiesioginės mazgo API palaikymas yra atskiri dalykai – vien naršyklės vaizdas neįrodo naujo forko taisyklių teisingumo.
