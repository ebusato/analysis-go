
void display(TString fileNameCenters="../../godaq/dpgageom.csv", TString fileNameFull="../../godaq/dpgafullgeom.csv", Bool_t register=kTRUE)
{
  TEveManager::Create();

  gStyle->SetPalette(1, 0);
  
  TEveRGBAPalette* pal = new TEveRGBAPalette(0, 130);
  
  // Read scintillators and put them in a box set
  TEveBoxSet* bs = new TEveBoxSet("BoxSet");
  bs->Reset(TEveBoxSet::kBT_FreeBox, kFALSE, 64);
  bs->SetPalette(pal);

  ifstream infull(fileNameFull.Data());
  if (!infull) {
    cerr << "ERROR ! Unable to open file '" << fileNameFull << "' !" << endl;
    return;
  }
  for(; !infull.eof() ;) {
    string line;
    if (!getline(infull,line)) break;
    if (!line.empty() && line[0]!='#') {
      istringstream istr(line);
      int iChannelAbs240;
      float X0, Y0, Z0;
      float X1, Y1, Z1;
      float X2, Y2, Z2;
      float X3, Y3, Z3;
      float X4, Y4, Z4;
      float X5, Y5, Z5;
      float X6, Y6, Z6;
      float X7, Y7, Z7;

      istr >> iChannelAbs240
	   >> X0 >> Y0 >> Z0
	   >> X1 >> Y1 >> Z1
	   >> X2 >> Y2 >> Z2
	   >> X3 >> Y3 >> Z3
	   >> X4 >> Y4 >> Z4
	   >> X5 >> Y5 >> Z5
	   >> X6 >> Y6 >> Z6
	   >> X7 >> Y7 >> Z7;
       
      Float_t verts[24] = {
	X0 , Y0 , Z0, 
	X1 , Y1 , Z1,
	X2 , Y2 , Z2,
	X3 , Y3 , Z3,
	X4 , Y4 , Z4,
	X5 , Y5 , Z5,
	X6 , Y6 , Z6,
	X7 , Y7 , Z7};
      bs->AddBox(verts);
      // Color code:
      //  10 -> Blue
      //  75 -> Green
      // 130 -> Red
      // if() {
      bs->DigitValue(100);
    }
  }
  bs->RefitPlex();

  // Read centers of scintillators' front faces and put them in a point set
  TEvePointSet* ps = new TEvePointSet();
  ps->SetOwnIds(kTRUE);
  
  ifstream incenters(fileNameCenters.Data());
  if (!incenters) {
    cerr << "ERROR ! Unable to open file '" << fileNameCenters << "' !" << endl;
    return;
  }
  for(; !incenters.eof() ;) {
    string line;
    if (!getline(incenters,line)) break;
    if (!line.empty() && line[0]!='#') {
      istringstream istr(line);
      int iChannelAbs240;
      float X, Y, Z;

      istr >> iChannelAbs240
	   >> X >> Y >> Z;
       
      ps -> SetNextPoint(X, Y, Z);
      ps -> SetPointId(new TNamed(Form("iChannelAbs240=%d", iChannelAbs240), ""));
    }
  }

  ps->SetMarkerColor(kBlue);
  ps->SetMarkerSize(2);
  ps->SetMarkerStyle(8);

  if (register)
    {
      gEve->AddElement(bs);
      gEve->AddElement(ps);
      gEve->Redraw3D(kTRUE);
    }
}
